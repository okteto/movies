var vote = "[(${vote})]";

  if(vote == "a"){
    $(".a").prop('disabled', true);
    $(".a").html('[(${optionA})] <i class="fa fa-check-circle"></i>');
    $(".b").css('opacity','0.5');
  }
  if(vote == "b"){
    $(".b").prop('disabled', true);
    $(".b").html('[(${optionB})] <i class="fa fa-check-circle"></i>');
    $(".a").css('opacity','0.5');
  }